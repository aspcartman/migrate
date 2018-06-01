package migrate

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	"github.com/go-pg/pg"
	"github.com/pkg/errors"
)

func MigrateFromFolder(db *pg.DB, path string, devmode bool) error {
	migs, err := GetMigrationsFromFolder(path)
	if err != nil {
		return err
	}
	return Migrate(db, migs, devmode)
}

func Migrate(db *pg.DB, migs []Migration, devmode bool) error {
	// Validate and prepare
	sort.Sort(migsSlice(migs)) // presort; that changes original data, but that should not be important
	err := validate(migs)
	if err != nil {
		return errors.Wrap(err, "validation error")
	}

	var tx *pg.Tx
	for i := 0; i < 15; i++ {
		if tx, err = db.Begin(); err != nil && i == 14 {
			return fmt.Errorf("tx start fail: %s", err.Error())
		} else if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	defer tx.Rollback()

	// Create migrations table
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS public.migrations (
		    id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			hash TEXT NOT NULL,
			date TIMESTAMP NOT NULL DEFAULT NOW()
		);
	`); err != nil {
		return fmt.Errorf("tx start fail: %s", err.Error())
	}

	// Acquire an explicit transactional lock on migrations
	if _, err := tx.Exec("select pg_advisory_lock(123)"); err != nil {
		return errors.Wrap(err, "acquiring lock")
	}

	// Fetch the whole migration state of the DB
	var state []struct {
		ID   int
		Name string
		Hash string
	}
	if _, err := tx.Query(&state, "select id, name, hash from migrations order by id asc"); err != nil && err != pg.ErrNoRows {
		return errors.Wrap(err, "state fetch")
	}

	// Validate the state
	if len(state) > len(migs) {
		return errors.New("db migration history has more migrations than provided array")
	}

	for i, s := range state {
		switch {
		case s.ID != migs[i].ID:
			return errors.Errorf("db migration state is inconsistent, %d != %d", s.ID, migs[i].ID)
		case s.Name != migs[i].Name:
			return errors.Errorf("db migration name does not equal the provided one for id=%d (%s != %s)", s.ID, s.Name, migs[i].Name)
		case s.Hash != hash(migs[i].Up) && (!devmode || i != len(state)-1):
			return errors.Errorf("db migration hash does not equal the provided one for id=%d (%s): %s != %s", s.ID, migs[i].Name, s.Hash, hash(migs[i].Up))
		}
	}

	if devmode && len(state) > 0 {
		if st, mg := state[len(state)-1], migs[len(migs)-1]; st.ID == mg.ID && st.Hash != hash(mg.Up) {
			fmt.Printf("devmode is on and migration %d (%s) changed, reverting\n", st.ID, st.Name)
			if _, err := tx.Exec(mg.Down); err != nil {
				return errors.Wrapf(err, "execution of migration DOWN script %d (%s) failed", mg.ID, mg.Name)
			}
			if _, err := tx.Exec(`delete from migrations where id = ?`, mg.ID); err != nil {
				return errors.Wrapf(err, "deleting row of migration script %d (%s) failed", mg.ID, mg.Name)
			}
			state = state[:len(state)-1]
		}
	}

	// Do migrations
	for _, mg := range migs[len(state):] {
		fmt.Println("applying", mg.ID, mg.Name)

		commands := SplitScript(mg.Up)
		if len(commands) == 0 {
			return errors.Errorf("empty up script %d (%s)", mg.ID, mg.Name)
		}
		for _, c := range commands {
			if _, err := tx.Exec(c); err != nil {
				return errors.Wrapf(err, "execution of migration script %d (%s) failed", mg.ID, mg.Name)
			}
		}
		if _, err := tx.Exec(`insert into public."migrations" (id,name,hash) values (?,?,?)`, mg.ID, mg.Name, hash(mg.Up)); err != nil {
			return errors.Wrapf(err, "saving execution of migration script %d (%s) failed", mg.ID, mg.Name)
		}
	}

	return tx.Commit()
}

func validate(migs []Migration) error {
	if len(migs) == 0 {
		return errors.New("empty migrations")
	}
	for i, m := range migs {
		switch {
		case m.ID != i+1:
			return errors.Errorf("inconsistant migrations versioning: expected %d got %d", i+1, m.ID)
		case len(m.Up) == 0:
			return errors.Errorf("empty UP script")
		case len(m.Down) == 0:
			return errors.Errorf("empty DOWN script")
		}
	}
	return nil
}

func hash(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

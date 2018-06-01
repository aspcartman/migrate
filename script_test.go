package migrate

import (
	"fmt"
	"strings"
	"testing"
)

func TestSplitScript(t *testing.T) {
	res := SplitScript(`
		--
		-- Name: ug_rel(integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
		--
		
		CREATE FUNCTION public.ug_rel(uid integer, gid integer) RETURNS void
		    LANGUAGE sql
		    AS $$
		update users set groups = groups || gid where id = uid;
		update groups set users = users || uid where id = gid;
		delete from user_groups ug where ug.user = uid and ug.group = gid;
		$$;
		
		
		ALTER FUNCTION public.ug_rel(uid integer, gid integer) OWNER TO postgres;
		
		--
		-- Name: ug_rel(tid, integer, integer); Type: FUNCTION; Schema: public; Owner: postgres
		--
	`)

	if len(res) != 2 {
		t.Errorf("wrong split result: %s", fmt.Sprint(res))
	}

	for i, stmt := range res {
		switch {
		case len(stmt) == 0:
			t.Errorf("empty stmtm #%d", i)
		case strings.HasPrefix(stmt, ";"):
			t.Errorf("stmt has ; :%s", stmt)
		case strings.HasSuffix(stmt, ";"):
			t.Errorf("stmt has ; :%s", stmt)
		}
	}

}

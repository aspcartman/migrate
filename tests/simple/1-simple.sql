CREATE TABLE IF NOT EXISTS public.users (
  id              INTEGER NOT NULL,
  name            TEXT    NOT NULL,
  birthdate       TEXT,
  location        TEXT,
  sex             INTEGER,
  fetched_groups  TIMESTAMP WITHOUT TIME ZONE,
  fetched_friends TIMESTAMP WITHOUT TIME ZONE,
  created_at      TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),
  groups          INTEGER []
);

-- DOWN

DROP TABLE public.users;
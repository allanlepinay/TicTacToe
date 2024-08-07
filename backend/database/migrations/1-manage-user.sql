ALTER TABLE players
ADD COLUMN password_hash TEXT;

ALTER TABLE players
ADD COLUMN wins INTEGER DEFAULT 0,
ADD COLUMN losses INTEGER DEFAULT 0,
ADD COLUMN draws INTEGER DEFAULT 0;
ADD CONSTRAINT unique_name UNIQUE (name);

CREATE INDEX idx_players_name ON players(name);

ALTER TABLE games
ADD COLUMN player_x_id INTEGER REFERENCES players(id),
ADD COLUMN player_o_id INTEGER REFERENCES players(id);

ALTER TABLE moves
DROP CONSTRAINT unique_move;

ALTER TABLE moves
ADD CONSTRAINT unique_move UNIQUE (game_id, x, y);
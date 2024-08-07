CREATE TABLE games (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) NOT NULL
);

CREATE TABLE moves (
    id SERIAL PRIMARY KEY,
    game_id INTEGER REFERENCES games(id) ON DELETE CASCADE,
    player CHAR(1) NOT NULL,
    x INTEGER NOT NULL,
    y INTEGER NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_move UNIQUE (game_id, x, y, timestamp)
);

CREATE TABLE players (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    score INTEGER DEFAULT 0
);

CREATE TABLE player_games (
    player_id INTEGER REFERENCES players(id) ON DELETE CASCADE,
    game_id INTEGER REFERENCES games(id) ON DELETE CASCADE,
    PRIMARY KEY (player_id, game_id)
);

CREATE INDEX idx_moves_game_id ON moves(game_id);
CREATE INDEX idx_moves_player ON moves(player);
CREATE INDEX idx_player_games_player_id ON player_games(player_id);
CREATE INDEX idx_player_games_game_id ON player_games(game_id);
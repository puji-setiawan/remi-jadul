-- migrate:up
CREATE TABLE completed_matches (
    match_id text PRIMARY KEY,
    round_count integer NOT NULL CHECK (round_count >= 1),
    completed_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE completed_match_players (
    match_id text NOT NULL REFERENCES completed_matches(match_id) ON DELETE CASCADE,
    player_id text NOT NULL,
    display_name text NOT NULL CHECK (length(btrim(display_name)) > 0),
    cumulative_score integer NOT NULL,
    rank integer NOT NULL CHECK (rank >= 1),
    PRIMARY KEY (match_id, player_id)
);

CREATE TABLE completed_round_scores (
    match_id text NOT NULL REFERENCES completed_matches(match_id) ON DELETE CASCADE,
    round_number integer NOT NULL CHECK (round_number >= 1),
    end_reason text NOT NULL,
    player_id text NOT NULL,
    score_status text NOT NULL,
    round_score integer,
    cumulative_score integer NOT NULL,
    PRIMARY KEY (match_id, round_number, player_id),
    FOREIGN KEY (match_id, player_id) REFERENCES completed_match_players(match_id, player_id) ON DELETE CASCADE
);

-- migrate:down
DROP TABLE IF EXISTS completed_round_scores;
DROP TABLE IF EXISTS completed_match_players;
DROP TABLE IF EXISTS completed_matches;

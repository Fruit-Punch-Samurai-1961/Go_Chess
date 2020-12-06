CREATE DATABASE chessapp;
USE chessapp;

#Table Layout
CREATE TABLE game
(
    id        INT PRIMARY KEY NOT NULL AUTO_INCREMENT,
    room      VARCHAR(255)    NOT NULL,
    fen       VARCHAR(255)    NOT NULL,
    canChange BOOLEAN         NOT NULL DEFAULT TRUE,
    expires   DATETIME        NOT NULL,
    CONSTRAINT room_unique UNIQUE (room)
);

#Insert Test
INSERT INTO game(room, fen, canChange, expires)
VALUES ('49a406a2-081f-11eb-adc1-0242ac120002', '2qb1k1N/3p4/N5B1/4R3/2p2P1p/2P5/n2K1P1P/8 w - - 0 1', TRUE,
        DATE_ADD(UTC_TIMESTAMP, INTERVAL 7 DAY));

#Select Test
SELECT (room)
FROM game
WHERE room = '49a406a2-081f-11eb-adc1-0242ac120002';

#Status Change Test
UPDATE game
SET canChange = TRUE
WHERE room = '49a406a2-081f-11eb-adc1-0242ac120002';

#Save Function
DELIMITER $$
CREATE FUNCTION save(room_given VARCHAR(255), newfen VARCHAR(255)) RETURNS BOOLEAN DETERMINISTIC
BEGIN
    SET @state = (SELECT canChange FROM game WHERE room = room_given);
    IF @state = TRUE THEN
        #Update canChange Status
        UPDATE game
        SET canChange = FALSE
        WHERE room = room_given;
        #Update the fen string
        UPDATE game
        SET fen = newfen
        WHERE room = room_given;
        #Update the expire date
        UPDATE game
        SET expires = DATE_ADD(UTC_TIME(), INTERVAL 7 DAY)
        WHERE room = room_given;
        RETURN TRUE;
    ELSE
        RETURN FALSE;
    END IF;
END $$

DELIMITER ;

#Save Test
SELECT save('5edc67c9-ba9e-424e-903d-18e2fb061225', '6k1/2P3b1/5P2/pP2NppK/4r3/p4Rp1/8/1b1Q4 w - - 0 1');


SELECT COUNT(*) FROM game;

DELETE FROM game WHERE id > 9;
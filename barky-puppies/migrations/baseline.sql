CREATE TABLE barky_puppies.`user` (
	id INT UNSIGNED auto_increment NOT NULL,
	displayName varchar(300) NOT NULL,
	apiKey varchar(300) NOT NULL,
	createdAt DATETIME NOT NULL,
	updatedAt DATETIME NOT NULL,
	CONSTRAINT user_PK PRIMARY KEY (id)
)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE barky_puppies.`user_last_synced` (
	userId INT UNSIGNED NOT NULL,
    lastSyncedAt DATE NOT NULL
    CONSTRAINT user_last_synced_PK PRIMARY KEY (userId) 
)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE barky_puppies.`user_stat_by_language` (
    userId INT UNSIGNED NOT NULL,
    `language` varchar(300) NOT NULL,
    recordedAt DATE NOT NULL,
    totalSeconds BIGINT NOT NULL,
    createdAt DATETIME NOT NULL,
	updatedAt DATETIME NOT NULL,
    CONSTRAINT user_last_synced_PK PRIMARY KEY (userId, language, recordedAt) 
)
ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_0900_ai_ci;

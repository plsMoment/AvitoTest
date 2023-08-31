CREATE TABLE "Segments" (
    "id" uuid PRIMARY KEY,
    "name" varchar NOT NULL
);

CREATE TABLE "UserSegments" (
    "user_id" uuid NOT NULL,
    "segment_id" uuid NOT NULL,
    FOREIGN KEY ("segment_id") REFERENCES "Segments" ("id"),
    PRIMARY KEY ("user_id", "segment_id")
);

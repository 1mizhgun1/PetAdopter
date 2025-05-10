CREATE TYPE ad_status_values AS ENUM ('A', 'R', 'C');

CREATE TABLE IF NOT EXISTS Region (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL CONSTRAINT region_name_length CHECK (char_length(name) <= 64)
);

CREATE TABLE IF NOT EXISTS Locality (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL CONSTRAINT locality_name_length CHECK (char_length(name) <= 64),
    latitude FLOAT NOT NULL,
    longitude FLOAT NOT NULL,
    region_id UUID NOT NULL REFERENCES Region (id)
);

CREATE TABLE IF NOT EXISTS Animal (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL CONSTRAINT animal_name_length CHECK (char_length(name) <= 64)
);

CREATE TABLE IF NOT EXISTS Breed (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL CONSTRAINT breed_name_length CHECK (char_length(name) <= 64),
    animal_id UUID NOT NULL REFERENCES Animal (id)
);

CREATE TABLE IF NOT EXISTS MyUser (
    id UUID PRIMARY KEY,
    username TEXT UNIQUE NOT NULL CONSTRAINT user_username_length CHECK (char_length(username) <= 20),
    password_hash TEXT NOT NULL CONSTRAINT user_password_hash_length CHECK (char_length(password_hash) <= 256),
    locality_id UUID REFERENCES Locality (id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE IF NOT EXISTS Ad (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES MyUser (id),
    status ad_status_values NOT NULL,
    photo_url TEXT CONSTRAINT ad_photo_url_length CHECK (char_length(photo_url) <= 128),
    title TEXT CONSTRAINT ad_title_length CHECK (char_length(title) <= 32),
    description TEXT CONSTRAINT ad_description_length CHECK (char_length(description) <= 4096),
    price INTEGER NOT NULL,
    animal_id UUID NOT NULL REFERENCES Animal (id),
    breed_id UUID NOT NULL REFERENCES Breed (id),
    contacts TEXT NOT NULL CONSTRAINT ad_contacts_length CHECK (char_length(contacts) <= 128),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE IF NOT EXISTS Favorite (
    user_id UUID REFERENCES MyUser (id),
    ad_id UUID REFERENCES Ad (id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (user_id, ad_id)
);

CREATE TABLE IF NOT EXISTS Watch (
    user_id UUID REFERENCES MyUser (id),
    ad_id UUID REFERENCES Ad (id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (user_id, ad_id)
);

CREATE TABLE IF NOT EXISTS History (
    user_id UUID PRIMARY KEY REFERENCES MyUser (id),
    animal_id UUID REFERENCES Animal (id),
    breed_id UUID REFERENCES Breed (id),
    min_price INTEGER,
    max_price INTEGER,
    radius INTEGER,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE IF NOT EXISTS GptDescription (
    id UUID PRIMARY KEY,
    color TEXT CONSTRAINT color_length CHECK (char_length(color) <= 32),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS locality_region_id_idx ON Locality (region_id);
CREATE INDEX IF NOT EXISTS breed_animal_id_idx ON Breed (animal_id);
CREATE INDEX IF NOT EXISTS ad_status_idx ON Ad (status);
CREATE INDEX IF NOT EXISTS ad_animal_id_idx ON Ad (animal_id);
CREATE INDEX IF NOT EXISTS ad_breed_id_idx ON Ad (breed_id);

CREATE OR REPLACE FUNCTION haversine_distance(
    lat1 FLOAT, lon1 FLOAT,
    lat2 FLOAT, lon2 FLOAT
) RETURNS FLOAT AS $$
DECLARE
    r CONSTANT FLOAT := 6371; -- Радиус Земли в километрах
    dlat FLOAT := radians(lat2 - lat1);
    dlon FLOAT := radians(lon2 - lon1);
    a FLOAT;
    c FLOAT;
    dist FLOAT;
BEGIN
    a := sin(dlat / 2)^2 + cos(radians(lat1)) * cos(radians(lat2)) * sin(dlon / 2)^2;
    c := 2 * atan2(sqrt(a), sqrt(1 - a));
    dist := r * c;

    -- Логирование входных и выходного значения
    RAISE NOTICE 'Haversine input: (%, %), (%, %) => dist: % km', lat1, lon1, lat2, lon2, dist;

    RETURN dist;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

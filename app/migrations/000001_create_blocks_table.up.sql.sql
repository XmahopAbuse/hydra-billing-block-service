BEGIN;

CREATE TABLE IF NOT EXISTS blocks (
    id serial PRIMARY KEY,
    customer_id varchar(200) NOT NULL,
    customer_code varchar(200) NOT NULL,
    customer_login varchar(200) NOT NULL,
    start_date date NOT NULL,
    end_date date DEFAULT NULL
    );

COMMIT;
END;
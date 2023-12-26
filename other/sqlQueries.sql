#postgres
create user wb_user;
grant all on schema public to wb_user;
#


#wb_user
CREATE TABLE delivery (
	id_delivery bigserial not null constraint PK_DELIVERY PRIMARY KEY,
    "name" VARCHAR(100),
    phone VARCHAR(20),
    zip VARCHAR(20),
    city VARCHAR(100),
    address VARCHAR(200),
    region VARCHAR(100),
    email VARCHAR(100)
);

CREATE TABLE payment (
	id_payment bigserial not null constraint pk_payment primary key,
    "transaction" VARCHAR(100),
    request_id VARCHAR(100),
    currency VARCHAR(10),
    provider VARCHAR(100),
    amount INTEGER,
    payment_dt INTEGER,
    bank VARCHAR(100),
    delivery_cost INTEGER,
    goods_total INTEGER,
    custom_fee INTEGER
);

CREATE TABLE items (
	id_items bigserial not null constraint pk_items primary key,
    order_id bigint not null referernces "order" (id_order)
    chrt_id INTEGER,
    track_number VARCHAR(100),
    price INTEGER,
    rid VARCHAR(100),
    name VARCHAR(100),
    sale INTEGER,
    size VARCHAR(10),
    total_price INTEGER,
    nm_id INTEGER,
    brand VARCHAR(100),
    status INTEGER
);

CREATE TABLE "order"(
	id_order BIGSERIAL NOT NULL CONSTRAINT PK_ORDER PRIMARY KEY,
	delivery_id bigint not null references delivery (id_delivery),
	payment_id bigint not null references payment (id_payment),
    order_uid VARCHAR(100),
    track_number VARCHAR(100),
    "entry" VARCHAR(100),
    locale VARCHAR(10),
    internal_signature VARCHAR(100),
    customer_id VARCHAR(100),
    delivery_service VARCHAR(100),
    shardkey VARCHAR(10),
    sm_id INTEGER,
    date_created TIMESTAMP,
    oof_shard VARCHAR(10)
);
#


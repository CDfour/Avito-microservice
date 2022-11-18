CREATE TABLE public.user
(
    id uuid PRIMARY KEY,
    balance decimal,
    date_create timestamp NOT NULL,
    last_update timestamp NOT NULL
);

CREATE TABLE public.order
(
    order_id uuid PRIMARY KEY,
    user_id uuid REFERENCES public.user(id),
    service_id uuid NOT NULL,
    service_name text NOT NULL,
    date_create date NOT NULL,
    funds decimal
);

CREATE TABLE public.accounting
(
    order_id uuid PRIMARY KEY,
    user_id uuid REFERENCES public.user(id),
    service_id uuid,
    service_name text NOT NULL,
    date_create date NOT NULL,
    funds decimal
);

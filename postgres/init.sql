-- Runs once on first init (docker-entrypoint-initdb.d)

CREATE USER user_svc    WITH PASSWORD 'user_secret';
CREATE USER product_svc WITH PASSWORD 'product_secret';
CREATE USER order_svc   WITH PASSWORD 'order_secret';

-- Each service has full rights in its own database
CREATE DATABASE users_db    OWNER user_svc;
CREATE DATABASE products_db OWNER product_svc;
CREATE DATABASE orders_db   OWNER order_svc;

-- revoke default connect so only each database's owner can connect
REVOKE CONNECT ON DATABASE users_db    FROM PUBLIC;
REVOKE CONNECT ON DATABASE products_db FROM PUBLIC;
REVOKE CONNECT ON DATABASE orders_db   FROM PUBLIC;

-- db/app_user.sql
-- KHUSUS DEVELOPMENT LOKAL. Password ini sengaja seragam di tim dan TIDAK dipakai di server.
-- Jalankan: mysql -u root -p < db/app_user.sql

CREATE USER IF NOT EXISTS 'sitangkap_app'@'localhost' IDENTIFIED BY 'sitangkap_dev';
CREATE USER IF NOT EXISTS 'sitangkap_app'@'127.0.0.1' IDENTIFIED BY 'sitangkap_dev';
GRANT SELECT, INSERT, UPDATE, DELETE ON sitangkap.* TO 'sitangkap_app'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON sitangkap.* TO 'sitangkap_app'@'127.0.0.1';
FLUSH PRIVILEGES;

-- db/schema.sql
-- Membuat database dan tabel SiTangkap. Aman dijalankan ulang (IF NOT EXISTS).
-- Jalankan: mysql -u root -p < db/schema.sql

CREATE DATABASE IF NOT EXISTS sitangkap
  CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE sitangkap;

CREATE TABLE IF NOT EXISTS users (
  id            INT UNSIGNED  NOT NULL AUTO_INCREMENT,
  username      VARCHAR(50)   NOT NULL,
  nama          VARCHAR(100)  NOT NULL,
  password_hash VARCHAR(100)  NOT NULL,
  role          ENUM('petugas','dinas') NOT NULL,
  aktif         TINYINT(1)    NOT NULL DEFAULT 1,
  dibuat_pada   TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uq_users_username (username)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS transaksi (
  id            INT UNSIGNED  NOT NULL AUTO_INCREMENT,
  nama_nelayan  VARCHAR(100)  NOT NULL,
  jenis_ikan    VARCHAR(60)   NOT NULL,
  berat_kg      DECIMAL(10,2) NOT NULL,
  harga_per_kg  INT UNSIGNED  NOT NULL,
  tanggal       DATE          NOT NULL,
  dibuat_oleh   INT UNSIGNED  NOT NULL,
  dibuat_pada   TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  diubah_pada   TIMESTAMP     NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_transaksi_tanggal (tanggal, id),
  CONSTRAINT fk_transaksi_user FOREIGN KEY (dibuat_oleh) REFERENCES users (id)
    ON DELETE RESTRICT ON UPDATE CASCADE,
  CONSTRAINT chk_transaksi_berat CHECK (berat_kg > 0 AND berat_kg <= 100000),
  CONSTRAINT chk_transaksi_harga CHECK (harga_per_kg > 0 AND harga_per_kg <= 100000000)
) ENGINE=InnoDB;

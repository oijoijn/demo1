-- もし my_app_db が存在しない場合のみ作成
CREATE DATABASE IF NOT EXISTS my_app_db;

-- my_app_db を使用
USE my_app_db;

-- users テーブルを作成
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- (オプション) 動作確認用の初期データ
INSERT INTO users (name, email) VALUES ('Taro Yamada', 'taro@example.com');
INSERT INTO users (name, email) VALUES ('Hanako Tanaka', 'hanako@example.com');

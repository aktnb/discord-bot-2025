-- bot_user ロールが存在しない場合は作成
DO $$ BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'bot_user') THEN
    CREATE ROLE bot_user;
  END IF;
END $$;

-- bot_user に voice_text_links テーブルへの権限を付与
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE voice_text_links TO bot_user;

-- 将来的なシーケンスやその他のオブジェクトへの権限も付与
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO bot_user;

-- 今後作成されるテーブルへのデフォルト権限も設定
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO bot_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO bot_user;

{ pkgs, lib, ... }: {
  channel = "stable-25.05";

  packages = [
 pkgs.go
    pkgs.nodejs_20
    pkgs.nodePackages.nodemon
  ];

  env = {
    PGHOST      = lib.mkForce "localhost";
    PGPORT      = "5432";
    PGDATABASE  = "mydatabase";
    PGUSER      = "myuser";
    PGPASSWORD  = "mypassword";
    DATABASE_URL = "postgresql://myuser:mypassword@localhost:5432/mydatabase?sslmode=disable";
  };

  services.postgres = {
    enable    = true;
    enableTcp = true;            # opción añadida en julio-24 :contentReference[oaicite:1]{index=1}
    package   = pkgs.postgresql_15;
    extensions = ["pgvector"];
    initialScript = pkgs.writeText "init.sql" ''
      DO $$
      BEGIN
        IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'myuser') THEN
          CREATE ROLE myuser LOGIN PASSWORD 'mypassword';
        END IF;
        IF NOT EXISTS (SELECT FROM pg_database WHERE datname = 'mydatabase') THEN
          CREATE DATABASE mydatabase OWNER myuser;
        END IF;
      END
      $$;
    '';
  };

  idx = {
    extensions = [ "golang.go" ];
    workspace.onCreate = {
      setupSchema = ''goose postgres "$DATABASE_URL" up'';
      default.openFiles = [ "backend/cmd/service/main.go" ];
    };
    previews = {
      enable   = true;
      previews = [
        {
          id = "api";
          env = env;
          command = [
            "bash" "-c" ''
              cd backend                          # ⬅️ contexto correcto
              go run ./cmd/service/main.go \
                -addr 0.0.0.0:$PORT
            ''
          ];
          manager = "web";
        }
        {
          id      = "web";
          env = {
            VITE_API_URL = "http://localhost:8080";
          };
          command = [
            "bash" "-c" ''
              cd frontend
              npm install
              npm run dev -- --port $PORT --host 0.0.0.0
            ''
          ];
          manager = "web";
        }
      ];
    };
  };
}


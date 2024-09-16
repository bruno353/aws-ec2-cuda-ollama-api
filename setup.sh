#!/bin/bash

# Verifica se o script está sendo executado como root
if [ "$EUID" -ne 0 ]
  then echo "Por favor, execute como root."
  exit
fi

# Variáveis de configuração
APP_USER="llmuser"
APP_DIR="/opt/llm-app"
DOMAIN_NAME="llm.techreport.ai"  # Substitua pelo seu domínio
API_KEY="sua_chave_de_api_aqui"

# Atualiza o sistema
apt update && apt upgrade -y

# Instala dependências
apt install -y golang-go nginx certbot

# Cria um usuário para o aplicativo
useradd -m -s /bin/bash "$APP_USER"

# Cria o diretório do aplicativo
mkdir -p "$APP_DIR"
chown "$APP_USER":"$APP_USER" "$APP_DIR"

# Clona o repositório do aplicativo
su - "$APP_USER" -c "
    cd $APP_DIR
    git clone https://github.com/bruno353/aws-ec2-cuda-ollama-api.git .
"

# Compila o aplicativo Go
su - "$APP_USER" -c "
    cd $APP_DIR
    go build -o llm-app
"

# Obtém o certificado SSL com Let's Encrypt
certbot certonly --standalone -d "$DOMAIN_NAME" --non-interactive --agree-tos --email blaureanosantos@gmail.com
CERT_FILE="/etc/letsencrypt/live/$DOMAIN_NAME/fullchain.pem"
KEY_FILE="/etc/letsencrypt/live/$DOMAIN_NAME/privkey.pem"

# Configura o serviço systemd
cat > /etc/systemd/system/llm-app.service <<EOL
[Unit]
Description=LLM Go Application Service
After=network.target

[Service]
Type=simple
User=$APP_USER
WorkingDirectory=$APP_DIR
ExecStart=$APP_DIR/llm-app
Environment=API_KEY=$API_KEY
Restart=always
LimitNOFILE=65536
AmbientCapabilities=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
EOL

# Recarrega o systemd e inicia o serviço
systemctl daemon-reload
systemctl enable llm-app.service
systemctl start llm-app.service

# Configura o Nginx como proxy reverso
cat > /etc/nginx/sites-available/llm-app.conf <<EOL
server {
    listen 80;
    server_name $DOMAIN_NAME;

    location /.well-known/acme-challenge/ {
        root /var/www/html;
    }

    location / {
        return 301 https://\$host\$request_uri;
    }
}

server {
    listen 443 ssl;
    server_name $DOMAIN_NAME;

    ssl_certificate $CERT_FILE;
    ssl_certificate_key $KEY_FILE;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection keep-alive;
        proxy_set_header Host \$host;
        proxy_cache_bypass \$http_upgrade;
    }
}
EOL

# Ativa a configuração do Nginx
ln -s /etc/nginx/sites-available/llm-app.conf /etc/nginx/sites-enabled/

# Testa a configuração do Nginx e reinicia o serviço
nginx -t && systemctl restart nginx

echo "Configuração concluída com sucesso!"

# AWS EC2 CUDA Ollama Setup

Forked from to ensure API security by inputting an ENV for the bearer auth, make sure to set the env on your OS system

## 1. Launch EC2 Instance

a. Go to EC2 dashboard and click "Launch instance"
b. Name your instance (e.g., "Ollama-GPU-Server")
c. AMI: Search for and select "Deep Learning Base OSS Nvidia Driver GPU AMI (Ubuntu 22.04)"
d. Instance type: g4dn.xlarge (4 vCPUs, 16 GiB Memory, 1 GPU)
e. Create or select a key pair for SSH access
f. Network settings: Create a security group with the following rules:
   - Allow SSH (port 22) from your IP
   - Allow Custom TCP (port 8080) from anywhere (0.0.0.0/0)
g. Configure storage: At least 30 GiB
h. Launch the instance

## 2. Connect to EC2 Instance

```
ssh -i your-key.pem ubuntu@your-ec2-public-ip
```

## 3. Update System

```
sudo apt update && sudo apt upgrade -y
```

## 4. Install Go

```
sudo apt install -y golang-go
```

## 5. Install Ollama

```
curl -fsSL https://ollama.com/install.sh | sh
```

## 6. Clone Repository

```
git clone https://github.com/developersdigest/aws-ec2-cuda-ollama-api.git
cd aws-ec2-cuda-ollama-api
```

## 7. Start Ollama in Background

```
ollama serve 
```

## 8. Pull Your Model

```
ollama pull gemma2:2b
```

## 10. Set Up Systemd Service - Required for setting up the ENV
# Either you can run export ```API_KEY="your_secure_api_key_here" ``` on your OS or set the env in this server script.

Create service file:
```
sudo vim /etc/systemd/system/ollama-api.service
```

Add the following content:
```
[Unit]
Description=Ollama API Service
After=network.target

[Service]
ExecStart=/usr/bin/go run /home/ubuntu/aws-ec2-cuda-ollama-api/main.go
WorkingDirectory=/home/ubuntu/aws-ec2-cuda-ollama-api
User=ubuntu
Environment=API_KEY=your_secure_api_key_here
Restart=always

[Install]
WantedBy=multi-user.target
```

Enable and start the service:
```
sudo systemctl enable ollama-api.service
sudo systemctl start ollama-api.service
```

## 11. Test API

From your local machine:
```
curl http://ec2-your-ec2.amazonaws.com:8080/v1/chat/completions \
-H "Content-Type: application/json" \
-H "Authorization: Bearer demo" \
-d '{
  "model": "gemma2:2b",
  "messages": [
    {"role": "user", "content": "Tell me a story about a brave knight"}
  ],
  "stream": true
}'
```

## Troubleshooting

- Ensure EC2 instance is running
- Verify Go application is running
- Check port 8080 is open in EC2 security group
- Confirm Ollama is running and "gemma2:2b" model is available

To check service status:
```
sudo systemctl status ollama-api.service
```

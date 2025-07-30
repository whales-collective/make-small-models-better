FROM debian:stable-slim

RUN <<EOF
apt-get update
apt-get install -y curl gpg jq
apt-get clean
rm -rf /var/lib/apt/lists/*
EOF

# Set working directory
WORKDIR /app

# Copy the main script
COPY main.sh /app/main.sh

# Start the main script
CMD ["bash", "/app/main.sh"]
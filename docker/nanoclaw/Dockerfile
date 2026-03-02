FROM node:20-slim

WORKDIR /app

RUN apt-get update && apt-get install -y \
    git \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

# Clone nanoclaw
RUN git clone https://github.com/qwibitai/nanoclaw.git . 2>/dev/null || true

RUN npm install --production

VOLUME ["/app/data"]

CMD ["node", "src/index.js"]

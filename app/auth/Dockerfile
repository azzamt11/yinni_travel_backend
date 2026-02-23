FROM node:20-alpine

WORKDIR /app

COPY app/auth/package*.json ./
RUN npm install --omit=dev

COPY app/auth ./

EXPOSE 8200
EXPOSE 9200
VOLUME /data/conf

CMD ["node", "index.js", "-conf", "/data/conf"]

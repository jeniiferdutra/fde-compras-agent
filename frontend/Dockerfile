FROM node:20-alpine

WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .

EXPOSE 5174

# Roda o Vite aceitando conexões de fora do container
CMD ["npm", "run", "dev", "--", "--host"]
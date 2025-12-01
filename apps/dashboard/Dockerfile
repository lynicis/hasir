FROM node:24-alpine

WORKDIR /app
COPY package*.json /.
RUN npm install

COPY . .
EXPOSE 3000
ENV NEXT_TELEMETRY_DISABLED=1
CMD npm run dev

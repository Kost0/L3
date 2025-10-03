FROM node:18-alpine as builder

WORKDIR /app

COPY *.html .
COPY *.css .
COPY *.js .

RUN mkdir -p /app/dist && \
    cp *.html /app/dist/ && \
    cp *.css /app/dist/

FROM node:18-alpine

WORKDIR /app

COPY --from=builder /app/dist/ .

RUN npm install -g http-server

EXPOSE 5000

CMD ["http-server", "-p", "5000"]
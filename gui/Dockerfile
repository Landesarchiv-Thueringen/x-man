FROM docker.io/node:alpine3.19 AS build

WORKDIR /build
COPY package.json package-lock.json ./
RUN npm ci
COPY angular.json tsconfig*.json ./
COPY src ./src
ARG CONFIGURATION
RUN npm run build -- --configuration $CONFIGURATION

FROM docker.io/nginx:alpine3.18 AS prod
COPY --from=build /build/dist/xdomea_appraisal_frontend /app/
COPY build/nginx/default.conf /etc/nginx/conf.d/default.conf

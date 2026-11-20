FROM nginx:1.29-alpine

# Refresh Alpine packages so the image ships current security fixes.
RUN apk upgrade --no-cache

# Remove default configuration
RUN rm -rf /etc/nginx/conf.d/* \
  && chown -R nginx:nginx /var/cache/nginx /var/log/nginx /etc/nginx/conf.d \
  && touch /var/run/nginx.pid \
  && chown nginx:nginx /var/run/nginx.pid

# Copy nginx config files
COPY nginx.conf /etc/nginx/nginx.conf
COPY includes/ /etc/nginx/includes/
COPY templates/ /etc/nginx/templates/

USER nginx

EXPOSE 80 443

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
  CMD wget -q --spider http://127.0.0.1/health || exit 1

# Start nginx
CMD ["nginx", "-g", "daemon off;"]

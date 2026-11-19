FROM nginx:1.27-alpine

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

# Start nginx
CMD ["nginx", "-g", "daemon off;"]

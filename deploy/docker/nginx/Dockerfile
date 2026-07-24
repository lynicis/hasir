FROM nginx:alpine

# Remove default configuration
RUN rm -rf /etc/nginx/conf.d/*

# Copy nginx config files
COPY nginx.conf /etc/nginx/nginx.conf
COPY includes/ /etc/nginx/includes/
COPY templates/ /etc/nginx/templates/

EXPOSE 80 443

# Start nginx
CMD ["nginx", "-g", "daemon off;"]

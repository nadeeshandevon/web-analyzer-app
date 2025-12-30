FROM nginx:alpine

# Copy the static files to the nginx html directory
COPY index.html /usr/share/nginx/html/
COPY app.js /usr/share/nginx/html/

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 80
EXPOSE 80

# Use a shell script to replace placeholders with environment variables at runtime
CMD ["/bin/sh", "-c", "sed -i \"s|API_BASE_URL_PLACEHOLDER|$API_BASE_URL|g\" /usr/share/nginx/html/app.js && sed -i \"s|API_KEY_PLACEHOLDER|$API_KEY|g\" /usr/share/nginx/html/app.js && nginx -g 'daemon off;'"]

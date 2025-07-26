FROM nginx:alpine

# Copy a simple HTML file
RUN echo '<html><body><h1>Hello from Docker Compose Build UI!</h1></body></html>' > /usr/share/nginx/html/index.html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
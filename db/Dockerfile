FROM redis:alpine

# Redis port
EXPOSE 6379

# Persist data inside container
VOLUME ["/data"]

CMD ["redis-server", "--appendonly", "yes"]

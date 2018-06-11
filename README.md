# whalepost
Deploy a new version of your docker service with a simple webhook.

* Registry authentication from Docker client config file
* Token to secure the deployment endpoint
* swarm label to allow update only for configured services

## Getting Started
Create the whalepost docker service. Make sure your docker client config is accessible by the executing swarm node.

    $: docker service create \
        --name=whalepost \
        --mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock \
        --mount type=bind,source=/root/.docker/config.json,target=/.docker/config.json \
        --constraint 'node.role == manager' \
        --publish 8000:8000 \
        faryon93/whalepost \
        /usr/sbin/whalrepost -token=s3cr3t

Use curl to test the webhook endpoint:

    $: curl -X POST -d '{"image": "jwilder/whoami", "auth": true}' http://localhost:8000/api/v1/service/test?key=s3cr3t

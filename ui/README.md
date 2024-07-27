# Architecture

This contains the html and javascript of the user interface. The user interface is a single page application that uses vite framework. It is build into static html/js and bundled for deployment

# Running the UI for development

Run the backend go server with yarn on your host:
    
    cd ui
    yarn backend-run

Bootstrap the database on the first run. Run the below init command in a new terminal, while the above run command is still running. Wait for the init command to finish. This step can be skipped on successive runs:

    cd ui
    yarn backend-init

When running on Linux, create a file at the path `ui/env.local` with the below content. The IP in the file should match your docker bridge. You can find the correct address with the command `ip address show`. On macOS you might be able to skip this step as the IP matches the contents of the `.env` file that is checked into the repo. 

    CLUSTER_MANAGER_BACKEND_IP=172.19.0.1

Install dotrun as described in https://github.com/canonical/dotrun#installation Launch it from the ui folder of this repo:

    cd ui
    dotrun

Now you can browse the ui via http://0.0.0.0:8414/

# End-to-end tests

Install playwright and its browsers

    npx playwright install

The tests expect the environment on localhost to be accessible. Execute `dotrun` first then run the tests against the latest LXD version with

    cd ui
    yarn test-e2e
# Architecture

This contains the html and javascript of the user interface. The user interface is a single page application that uses vite framework. It is build into static html/js and bundled for deployment

# Building the UI

Install dotrun as described in https://github.com/canonical/dotrun#installation Launch it from the ui folder of this repo

    cd ui
    dotrun

# End-to-end tests

Install playwright and its browsers

    npx playwright install

The tests expect the environment on localhost to be accessible. Execute `dotrun` first then run the tests against the latest LXD version with

    cd ui
    yarn test-e2e
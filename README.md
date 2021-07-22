# shipwright-poc

Steps to run:

1. Start your own k8s cluster (like k3d for eg) and make sure that your current context is set to that cluster 
2. Execute ```go run *.go``` from within the ```shipwright-poc``` folder 
3. Navigate to ```localhost:8080/form.html``` and fill out the required details and hit submit (quay support is still WIP). Wait for it to finish processing and update the page
4. Once build is successful, navigate to your image registry to see the pushed image 

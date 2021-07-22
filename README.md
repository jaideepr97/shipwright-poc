# shipwright-poc

Steps to run:

1. Edit ```server.go > kubeConfigPath``` to point to your ``.kube/config``` file
2. Start your own k8s cluster (like k3d for eg) and make sure that your current context is set to that cluster 
3. Execute ```go run *.go``` from within the ```shipwright-poc``` folder 
4. Navigate to ```localhost:8080/form.html``` and fill out the required details and hit submit (quay support is still WIP)
5. Once build is successful, navigate to your image registry to see the pushed image 

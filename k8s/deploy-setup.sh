echo "starting minikube"
minikube start

echo "enabling metrics addon"
minikube addons enable metrics-server

echo "loading images"
minikube image load postgres:16-alpine
minikube image load trainee-go-cart-api-app:latest

echo "creating namespace"
kubectl apply -f k8s/test/namespace/test.yaml

echo "applying manifests"
kubectl apply -R -f k8s/test

echo "created resources:"
kubectl get pods,svc,hpa,pvc -n test

echo "To work with app run 'kubectl port-forward service/cart-api-nodeport 8080:80'"
For Kafka Ca Issuer

    ```openssl req -newkey rsa:2048 -keyout ca.key -nodes -x509 -days 3650 -out ca.crt```
    ```kubectl create secret tls kafka-ca --cert=ca.crt --key=ca.key --namespace=demo```

For Kafka New Ca Issuer

    ```openssl req -newkey rsa:2048 -keyout ca-new.key -nodes -x509 -days 3650 -out ca-new.crt```
    ```kubectl create secret tls kafka-ca-new --cert=ca-new.crt --key=ca-new.key --namespace=demo```


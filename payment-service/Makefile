tls:
	mkdir -p certs

	openssl genrsa -out certs/ca.key 4096
	openssl req -new -x509 -days 3650 -key certs/ca.key -out certs/ca.crt -subj "/CN=CSFLOW-CA"

	echo "subjectAltName=DNS:payment-service,DNS:localhost,IP:127.0.0.1" > certs/payment.ext
	echo "extendedKeyUsage=serverAuth" >> certs/payment.ext

	openssl req -new -newkey rsa:2048 -nodes \
		-keyout certs/payment-service.key \
		-out certs/payment-service.csr \
		-subj "/CN=payment-service"
	openssl x509 -req -in certs/payment-service.csr \
		-CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial \
		-out certs/payment-service.crt -days 365 \
		-extfile certs/payment.ext

	rm -f ./certs/*.csr certs/*.srl certs/*.ext


tls-client:
	openssl genrsa -out ./certs/client.key 2048
	openssl req -new -key ./certs/client.key -subj "/CN=test-client" -out ./certs/client.csr
	openssl x509 -req -in ./certs/client.csr -CA ./certs/ca.crt -CAkey ./certs/ca.key -CAcreateserial -out ./certs/client.crt -days 365

	rm -f ./certs/*.csr *.srl
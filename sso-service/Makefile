tests:
	go test ./internal/infrastructure/config/tests/config_test.go
	go test ./internal/domain/user/tests/user_test.go

keys:
	openssl genrsa -out jwtRS256.key 2048
	openssl rsa -in jwtRS256.key -pubout -out jwtRS256.key.pub

tls:
	mkdir -p certs

	openssl genrsa -out certs/ca.key 4096
	openssl req -new -x509 -days 3650 -key certs/ca.key -out certs/ca.crt -subj "/CN=CSFLOW-CA"

	echo "subjectAltName=DNS:sso-service,DNS:localhost,IP:127.0.0.1" > certs/sso.ext
	echo "extendedKeyUsage=serverAuth" >> certs/sso.ext

	openssl req -new -newkey rsa:2048 -nodes \
		-keyout certs/sso-service.key \
		-out certs/sso-service.csr \
		-subj "/CN=sso-service"
	openssl x509 -req -in certs/sso-service.csr \
		-CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial \
		-out certs/sso-service.crt -days 365 \
		-extfile certs/sso.ext

	rm -f ./certs/*.csr certs/*.srl certs/*.ext


tls-client:
	openssl genrsa -out ./certs/client.key 2048
	openssl req -new -key ./certs/client.key -subj "/CN=test-client" -out ./certs/client.csr
	openssl x509 -req -in ./certs/client.csr -CA ./certs/ca.crt -CAkey ./certs/ca.key -CAcreateserial -out ./certs/client.crt -days 365

	rm -f ./certs/*.csr *.srl
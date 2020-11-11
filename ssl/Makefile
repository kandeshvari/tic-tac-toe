.phony: ca service clean

all: service

rootCA.key rootCA.crt:
		./gencerts.sh ca

ca: rootCA.key rootCA.crt

cert.pem key.pem: rootCA.key rootCA.crt
		./gencerts.sh service

service: cert.pem key.pem

clean:
		./gencerts.sh clean


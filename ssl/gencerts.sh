#!/bin/bash

export OPENSSL_CONF=./openssl.conf

host="localhost"
ipaddr="127.0.0.1"

script_dir="${0%%/*}"

usage() {
	echo "usage: ${0##*/} ca | service -h <host> -a <ip_addr> | ca-show | service-show"
	echo ""
	echo "Commands:"
	echo "  ca - generate CA cert"
	echo ""
	echo "  service - generate service cert"
	echo "      host        alternative dns name for cert (default: localhost)"
	echo "      ipaddr      alternative ip for cert (default: 127.0.0.1)"
	echo ""
	echo "  ca-show - show CA cert"
	echo ""
	echo "  service-show - show service cert"
	echo ""
	echo "  clean - remove all certs, requests and other generated artefacts"
	echo ""
	exit 1
}

CASubj="/C=FI/ST=Pajat-Hame /L=Lahti /O=tic-tac-toe corp./OU=sd/CN=tic-tac-toe CA/emailAddress=ca@tic-tac-toe.ltd"
SRVSubj="/C=FI/ST=Pajat-Hame /L=Lahti /O=tic-tac-toe corp./OU=sd/CN=tic-tac-toe web service/emailAddress=web@tic-tac-toe.tld"

pushd $script_dir >/dev/null

case $1 in
	ca)

		openssl req -x509 -new -nodes -sha512 -days 60 -newkey rsa:4096 -keyout rootCA.key -out rootCA.crt \
				 -subj "${CASubj}" -extensions v3_ca
		;;

	service)
		shift
		while getopts "a:h:" opt; do
    			case $opt in
    				a) ipaddr=$OPTARG ;;
				h) host=$OPTARG ;;
				*) usage ;;
    			esac
    		done

		openssl req -new -nodes -out service.csr -newkey rsa:4096 -keyout key.pem \
			-subj "${SRVSubj}"
		openssl x509 -req -in service.csr \
			-CA rootCA.crt -CAkey rootCA.key -CAcreateserial \
			-out cert.pem -days 30 -sha512 \
			-extfile <(cat ./v3.ext <(printf "subjectAltName = DNS:${host},IP:${ipaddr}\n"))
		;;

	clean)
		rm -f *.crt *.csr *.srl *.pem *.key
		;;

	ca-show)
		openssl x509 -in rootCA.crt -text -noout
		;;

	service-show)
		openssl x509 -in cert.pem -text -noout
		;;

	*)
		usage
		;;

esac
popd >/dev/null



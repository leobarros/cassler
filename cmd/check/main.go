package check

import (
	"crypto/tls"
	"crypto/x509/pkix"
	"fmt"
	"net"
	"time"

	"github.com/msfidelis/cassler/src/libs/lookup"
	"github.com/msfidelis/cassler/src/libs/parser"
)

type Certificate struct {
	CommonName            string
	NotBefore             time.Time
	NotAfter              time.Time
	TimeRemain            time.Duration
	IssuingCertificateURL []string
	SignatureAlgorithm    string
	Version               int
	Issuer                pkix.Name
	Subject               pkix.Name
	DNSNames              []string
}

func Cmd(url string, port int, dns_server string) {

	host := parser.ParseHost(url)
	ips := lookup.Lookup(host, dns_server)

	checked_certificates := make(map[string]string)
	certificate_authorities := make(map[string]Certificate)
	certificate_list := make(map[string]Certificate)

	fmt.Printf("Checking Certificates: %s on port %d \n", host, port)
	fmt.Printf("\nDNS Lookup on: %s \n\n", dns_server)

	for _, ip := range ips {

		conn_config := &tls.Config{
			ServerName:         host,
			InsecureSkipVerify: true,
		}

		dialer := net.Dialer{Timeout: 1000000000, Deadline: time.Now().Add(1000000000 + 5*time.Second)}
		connection, err := tls.DialWithDialer(&dialer, "tcp", fmt.Sprintf("[%s]:%d", ip, port), conn_config)

		if err != nil {
			fmt.Printf("%v\n", err)
		} else {

			certificate_negotiation_list := connection.ConnectionState().PeerCertificates

			for i := 0; i < len(certificate_negotiation_list); i++ {
				cert := certificate_negotiation_list[i]

				// Filter Certificate Already Validated
				if _, checked := checked_certificates[string(cert.Signature)]; checked {
					continue
				}

				checked_certificates[string(cert.Signature)] = cert.Subject.CommonName

				var certificate Certificate

				certificate.CommonName = cert.Subject.CommonName
				certificate.NotAfter = cert.NotAfter
				certificate.NotBefore = cert.NotBefore
				certificate.TimeRemain = cert.NotAfter.Sub(time.Now())
				certificate.SignatureAlgorithm = cert.SignatureAlgorithm.String()
				certificate.IssuingCertificateURL = cert.IssuingCertificateURL
				certificate.Version = cert.Version
				certificate.DNSNames = cert.DNSNames
				certificate.Issuer = cert.Issuer
				certificate.Subject = cert.Subject

				// Filter Certificate Authority
				if cert.IsCA {
					certificate_authorities[string(cert.Subject.CommonName)] = certificate
					continue
				}

				certificate_list[string(cert.Subject.CommonName)] = certificate

			}
		}

	}

	fmt.Printf("Server Certificate: \n")
	for _, data := range certificate_list {
		fmt.Printf("Common Name: %s\n", data.CommonName)
		fmt.Printf("Issuer: %s\n", data.Issuer)
		fmt.Printf("Subject: %s\n", data.Subject)
		fmt.Printf("Signature Algorithm: %s\n", data.SignatureAlgorithm)
		fmt.Printf("Created: %s\n", data.NotBefore)
		fmt.Printf("Expires: %s\n", data.NotAfter)
		fmt.Printf("Expiration time: %d days\n", parser.ParseDurationInDays(data.TimeRemain.Hours()))
		fmt.Printf("Certificate Version: %d\n", data.Version)

		if len(data.DNSNames) > 0 {
			fmt.Printf("\nDNS Names: \n")
			for _, dns := range data.DNSNames {
				fmt.Printf("- %s\n", dns)
			}
		}

		if len(data.IssuingCertificateURL) > 0 {
			fmt.Printf("\nIssuing Certificate URL's: \n")
			for _, url := range data.IssuingCertificateURL {
				fmt.Printf("- %s\n", url)
			}
		}
	}

	fmt.Printf("\nServer IP's: \n")
	for _, ip := range ips {
		fmt.Printf("* %s \n", ip)
	}

	fmt.Printf("\nCertificate Authority: \n\n")
	for _, data := range certificate_authorities {

		fmt.Printf("%s\n", data.CommonName)
		fmt.Printf("Issuer: %s\n", data.Issuer)
		fmt.Printf("Subject: %s\n", data.Subject)
		fmt.Printf("Signature Algorithm: %s\n", data.SignatureAlgorithm)
		fmt.Printf("Created: %s\n", data.NotBefore)
		fmt.Printf("Expires: %s\n", data.NotAfter)
		fmt.Printf("Expiration time: %d days\n", parser.ParseDurationInDays(data.TimeRemain.Hours()))
		fmt.Printf("Certificate Version: %d\n", data.Version)

		if len(data.IssuingCertificateURL) > 0 {
			fmt.Printf("\n\nIssuing Certificate URL's: \n")
			for _, url := range data.IssuingCertificateURL {
				fmt.Printf("- %s\n", url)
			}
		}

		fmt.Printf("\n\n")
	}

}

package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/Lyruik/gosipProject/helpers/sipDatabaseHelpers"
	"github.com/icholy/digest"
)

func main() {
	extIP := flag.String("ip", "10.12.1.50:5060", "My external ip")
	creds := flag.String("u", "alice:alice", "Comma seperated username:password list")
	tran := flag.String("t", "udp", "Transport")
	tlskey := flag.String("tlskey", "", "TLS key path")
	tlscrt := flag.String("tlscrt", "", "TLS crt path")
	flag.Parse()

	sip.SIPDebug = os.Getenv("SIP_DEBUG") != ""

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro
	log.Logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.StampMicro,
	}).With().Timestamp().Logger().Level(zerolog.InfoLevel)

	if lvl, err := zerolog.ParseLevel(os.Getenv("LOG_LEVEL")); err == nil && lvl != zerolog.NoLevel {
		log.Logger = log.Logger.Level(lvl)
	}

	registry := databaser.PullRegistry()
	/*for _, c := range strings.Split(*creds, ",") {
		arr := strings.Split(c, ":")
		registry[arr[0]] = arr[1]
	}*/

	ua, err := sipgo.NewUA(
		sipgo.WithUserAgent("SIPGO"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Fail to setup user agent")
	}

	srv, err := sipgo.NewServer(ua)
	if err != nil {
		log.Fatal().Err(err).Msg("Fail to setup server handle")
	}

	ctx := context.TODO()

	var chal digest.Challenge
	srv.OnRegister(func(req *sip.Request, tx sip.ServerTransaction) {
		h := req.GetHeader("Authorization")
		if h == nil {
			chal = digest.Challenge{
				Realm:     "sipgo-server",
				Nonce:     fmt.Sprintf("%d", time.Now().UnixMicro()),
				Opaque:    "sipgo",
				Algorithm: "MD5",
			}

			res := sip.NewResponseFromRequest(req, 401, "Unauthorized", nil)
			res.AppendHeader(sip.NewHeader("WWW-Authenticate", chal.String()))

			tx.Respond(res)
			return
		}

		cred, err := digest.ParseCredentials(h.Value())
		if err != nil {
			log.Error().Err(err).Msg("parsing creds failed")
			tx.Respond(sip.NewResponseFromRequest(req, 401, "Bad credentials", nil))
			return
		}

		passwd, exists := registry[cred.Username]
		if !exists {
			tx.Respond(sip.NewResponseFromRequest(req, 404, "Bad authorization header", nil))
			return
		}

		digCred, err := digest.Digest(&chal, digest.Options{
			Method:   "REGISTER",
			URI:      cred.URI,
			Username: cred.Username,
			Password: passwd,
		})

		if err != nil {
			log.Error().Err(err).Msg("Calc digest failed")
			tx.Respond(sip.NewResponseFromRequest(req, 401, "Bad Credentials", nil))
			return
		}

		if cred.Response != digCred.Response {
			tx.Respond(sip.NewResponseFromRequest(req, 401, "Unauthorized", nil))
			return
		}

		log.Info().Str("username", cred.Username).Str("source", req.Source()).Msg("New client registered")
		tx.Respond(sip.NewResponseFromRequest(req, 200, "OK", nil))
	})

	log.Info().Str("addr", *extIP).Msg("listening on")

	switch *tran {
	case "tls", "wss":
		cert, err := tls.LoadX509KeyPair(*tlscrt, *tlskey)
		if err != nil {
			log.Fatal().Err(err).Msg("Fail to load x509 key and crt")
		}
		if err := srv.ListenAndServeTLS(ctx, *tran, *extIP, &tls.Config{Certificates: []tls.Certificate{cert}}); err != nil {
			log.Info().Err(err).Msg("listening stop")
		}
		return
	}
	srv.ListenAndServe(ctx, *tran, *extIP)
}

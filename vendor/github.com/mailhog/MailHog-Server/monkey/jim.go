package monkey

import (
	"flag"
	"math/rand"
	"net"
	"time"

	"github.com/ian-kent/linkio"
)

// Jim is a chaos monkey
type Jim struct {
	DisconnectChance      float64
	AcceptChance          float64
	LinkSpeedAffect       float64
	LinkSpeedMin          float64
	LinkSpeedMax          float64
	RejectSenderChance    float64
	RejectRecipientChance float64
	RejectAuthChance      float64
	logf                  func(message string, args ...interface{})
}

// RegisterFlags implements ChaosMonkey.RegisterFlags
func (j *Jim) RegisterFlags() {
	flag.Float64Var(&j.DisconnectChance, "jim-disconnect", 0.005, "Chance of disconnect")
	flag.Float64Var(&j.AcceptChance, "jim-accept", 0.99, "Chance of accept")
	flag.Float64Var(&j.LinkSpeedAffect, "jim-linkspeed-affect", 0.1, "Chance of affecting link speed")
	flag.Float64Var(&j.LinkSpeedMin, "jim-linkspeed-min", 1024, "Minimum link speed (in bytes per second)")
	flag.Float64Var(&j.LinkSpeedMax, "jim-linkspeed-max", 10240, "Maximum link speed (in bytes per second)")
	flag.Float64Var(&j.RejectSenderChance, "jim-reject-sender", 0.05, "Chance of rejecting a sender (MAIL FROM)")
	flag.Float64Var(&j.RejectRecipientChance, "jim-reject-recipient", 0.05, "Chance of rejecting a recipient (RCPT TO)")
	flag.Float64Var(&j.RejectAuthChance, "jim-reject-auth", 0.05, "Chance of rejecting authentication (AUTH)")
}

// Configure implements ChaosMonkey.Configure
func (j *Jim) Configure(logf func(string, ...interface{})) {
	j.logf = logf
	rand.Seed(time.Now().Unix())
}

// ConfigureFrom lets us configure a new Jim from an old one without
// having to expose logf (and any other future private vars)
func (j *Jim) ConfigureFrom(j2 *Jim) {
	j.Configure(j2.logf)
}

// Accept implements ChaosMonkey.Accept
func (j *Jim) Accept(conn net.Conn) bool {
	if rand.Float64() > j.AcceptChance {
		j.logf("Jim: Rejecting connection\n")
		return false
	}
	j.logf("Jim: Allowing connection\n")
	return true
}

// LinkSpeed implements ChaosMonkey.LinkSpeed
func (j *Jim) LinkSpeed() *linkio.Throughput {
	rand.Seed(time.Now().Unix())
	if rand.Float64() < j.LinkSpeedAffect {
		lsDiff := j.LinkSpeedMax - j.LinkSpeedMin
		lsAffect := j.LinkSpeedMin + (lsDiff * rand.Float64())
		f := linkio.Throughput(lsAffect) * linkio.BytePerSecond
		j.logf("Jim: Restricting throughput to %s\n", f)
		return &f
	}
	j.logf("Jim: Allowing unrestricted throughput")
	return nil
}

// ValidRCPT implements ChaosMonkey.ValidRCPT
func (j *Jim) ValidRCPT(rcpt string) bool {
	if rand.Float64() < j.RejectRecipientChance {
		j.logf("Jim: Rejecting recipient %s\n", rcpt)
		return false
	}
	j.logf("Jim: Allowing recipient%s\n", rcpt)
	return true
}

// ValidMAIL implements ChaosMonkey.ValidMAIL
func (j *Jim) ValidMAIL(mail string) bool {
	if rand.Float64() < j.RejectSenderChance {
		j.logf("Jim: Rejecting sender %s\n", mail)
		return false
	}
	j.logf("Jim: Allowing sender %s\n", mail)
	return true
}

// ValidAUTH implements ChaosMonkey.ValidAUTH
func (j *Jim) ValidAUTH(mechanism string, args ...string) bool {
	if rand.Float64() < j.RejectAuthChance {
		j.logf("Jim: Rejecting authentication %s: %s\n", mechanism, args)
		return false
	}
	j.logf("Jim: Allowing authentication %s: %s\n", mechanism, args)
	return true
}

// Disconnect implements ChaosMonkey.Disconnect
func (j *Jim) Disconnect() bool {
	if rand.Float64() < j.DisconnectChance {
		j.logf("Jim: Being nasty, kicking them off\n")
		return true
	}
	j.logf("Jim: Being nice, letting them stay\n")
	return false
}

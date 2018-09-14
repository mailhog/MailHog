Deploying MailHog
=================

### Command line

You can run MailHog locally from the command line.

    go get github.com/mailhog/MailHog
    MailHog -h

To configure MailHog, use the environment variables or command line flags
described in the [CONFIG](CONFIG.md).

### Using supervisord/upstart/etc

MailHog can be started as a daemon using supervisord/upstart/etc.

See [this example init script](https://github.com/geerlingguy/ansible-role-mailhog/blob/master/templates/mailhog.init.j2)
and [this Ansible role](https://github.com/geerlingguy/ansible-role-mailhog) by [geerlingguy](https://github.com/geerlingguy).

If installed with Homebrew on OSX you can have launchd start mailhog now and restart at login:
    brew services start mailhog

### Docker

The example [Dockerfile](../Dockerfile) can be used to run MailHog in a [Docker](https://www.docker.com/) container.

You can run it directly from Docker Hub (thanks [humboldtux](https://github.com/humboldtux))

    docker run -d -p 1025:1025 -p 8025:8025 mailhog/mailhog

To mount the Maildir to the local filesystem, you can use a volume:

    docker run -d -e "MH_STORAGE=maildir" -v $PWD/maildir:/maildir -p 1025:1025 -p 8025:8025 mailhog/mailhog

### Elastic Beanstalk

You can deploy MailHog using [AWS Elastic Beanstalk](http://aws.amazon.com/elasticbeanstalk/).

1. Open the Elastic Beanstalk console
2. Create a zip file containing the Dockerfile and MailHog binary
3. Create a new Elastic Beanstalk application
4. Launch a new environment and upload the zip file

**Note** You'll need to reconfigure nginx in Elastic Beanstalk to expose both
ports as TCP, since by default it proxies the first exposed port to port 80 as HTTP.

If you're using in-memory storage, you can only use a single instance of
MailHog. To use a load balanced EB application, use MongoDB backed storage.

To configure your Elastic Beanstalk MailHog instance, either:

* Set environment variables using the Elastic Beanstalk console
* Edit the Dockerfile to pass in command line arguments

You may face restrictions on outbound SMTP from EC2, for example if you are
releasing messages to real SMTP servers.

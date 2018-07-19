FROM scratch

ADD bin/eve-axiom /

ENTRYPOINT ["/eve-axiom"]

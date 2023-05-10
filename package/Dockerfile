FROM registry.suse.com/bci/bci-micro:15.4

ARG user=adapter

RUN echo "$user:x:1000:1000::/home/$user:/bin/bash" >> /etc/passwd && \
    echo "$user:x:1000:" >> /etc/group && \
    mkdir /home/$user && \
    chown -R $user:$user /home/$user

COPY bin/csp-adapter /usr/bin/

USER $user

CMD ["csp-adapter"]

FROM python:3.7-alpine

RUN apk --no-cache add socat
RUN apk add gcc g++ make libffi-dev openssl-dev
RUN pip3 install pycryptodome

COPY ./prob.py /prob.py

CMD socat TCP-L:8888,fork,reuseaddr EXEC:"python prob.py",

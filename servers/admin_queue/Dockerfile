FROM python:3.6.5-slim
WORKDIR /app
ADD . /app
RUN pip3 install redis
RUN pip3 install pika
RUN pip3 install Flask-PyMongo
RUN pip3 install --trusted-host pypi.python.org -r requirements.txt
EXPOSE 80
ENTRYPOINT [ "python", "admin.py" ]
CMD [ "admin.py" ]
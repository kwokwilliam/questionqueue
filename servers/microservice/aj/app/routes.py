from app import app
from flask import request, Response
from bson.objectid import ObjectId
from flask_pymongo import PyMongo
import pymongo
import os
import json
import datetime
import redis
import pika

# Constants and environment variables
JSON_TYPE = 'application/json'
TEXT_TYPE = 'text/plain'

MONGO_URI = os.getenv(
    "MONGO_URI", "mongodb://localhost:27017/question_queue")
REDIS_HOST = os.getenv("REDIS_HOST", 'localhost')
REDIS_PORT = os.getenv("REDIS_PORT", 6379)
QUEUE_NAME = os.getenv("QUEUE_NAME", 'queue')
RABBIT_HOST = os.environ.get('RABBIT_HOST', "localhost")

# MongoDB configuration
app.config["MONGO_URI"] = MONGO_URI
mongo = PyMongo(app)
db = mongo.db

classes = os.getenv("CLASS_COLLECTION", 'class')
question = os.getenv("QUESTION_COLLECTION", 'question')

# Redis configuration
r = redis.StrictRedis(
    host=REDIS_HOST,
    port=REDIS_PORT,
    password='')

# RabbitMQ configuration
params = pika.ConnectionParameters(host=RABBIT_HOST, heartbeat=0)
connection = pika.BlockingConnection(params)
mq_channel = connection.channel()
mq_channel.queue_declare(queue=QUEUE_NAME, durable=True)


# GET a current class or POST a new class
@app.route('/v1/class', methods=['GET', 'POST'])
def class_handler():

    # RW
    print("inside class handler")

    if request.method == 'GET':
        all_classes = []
        try:
            all_classes = list(db[classes].find())
            for c in all_classes:
                c['_id'] = str(c['_id'])

            # added by RW for debugging
            # not filling class info to the client
            print("inspecting all classes")
            print(all_classes)
        except pymongo.errors.PyMongoError:
            return handle_db_error()

        resp = Response(json.dumps(all_classes), status=200,
                        mimetype=JSON_TYPE)
        return resp

    elif request.method == 'POST':
        # Check for authentication
        auth = check_auth(request)
        if auth != None:
            return auth

        # Check content type
        content = check_content_type(request)
        if content != None:
            return content

        req_body = request.get_json()

        if req_body.get("class_number", "") == "":
            return handle_missing_field("Class number is required")
        elif req_body.get("topics", "") == "" or not isinstance(req_body['topics'], list):
            return handle_missing_field("Class topics are required")

        # if not isinstance(req_body['class_number'], int):
        #     resp = Response("Class number must be an integer",
        #                     status=400, mimetype=TEXT_TYPE)
        #     return resp

        # Check if it already exists in the database
        class_query = {"class_number": req_body['class_number']}
        class_check, err = check_for_object(class_query, 'class')
        if class_check != None:
            return err

        # Insert into database
        new_class = {
            "class_number": req_body['class_number'],
            "topics": req_body['topics']
        }

        try:
            db[classes].insert_one(new_class)
        except pymongo.errors.PyMongoError:
            return handle_db_error()

        new_class['_id'] = str(new_class['_id'])

        resp = Response(json.dumps(new_class), status=201,
                        mimetype=JSON_TYPE)
        return resp


# PATCH an existing class - overwrite topics
@app.route('/v1/class/<class_number>', methods=['PATCH'])
def specific_class_handler(class_number):
    # Check for authentication
    auth = check_auth(request)
    if auth != None:
        return auth

    if request.method == 'PATCH':
        # Check content type
        content = check_content_type(request)
        if content != None:
            return content

        # Check if the requested class exists in the database
        class_query = {"class_number": class_number}
        req_class, err = check_for_object(class_query, 'class')
        if req_class == None:
            return err

        # Retrieve JSON body, check for validity, and update
        req_body = request.get_json()
        if req_body.get("topics", "") == "" or not isinstance(req_body['topics'], list):
            return handle_missing_field("Class topics are required")

        topics = list(req_body['topics'])
        update_query = {"$set": {"topics": topics}}

        updated = {}
        try:
            db[classes].update_one(class_query, update_query)
            updated = db[classes].find_one(class_query)
            updated['_id'] = str(updated['_id'])
        except pymongo.errors.PyMongoError:
            return handle_db_error()

        resp = Response(json.dumps(updated), status=200, mimetype=JSON_TYPE)
        return resp


# DELETE a student and question from the queue
@app.route('/v1/queue/<student_id>', methods=['DELETE'])
def queue_delete_handler(student_id):
    if request.method == 'DELETE':
        # Update resolution in mongo
        q_query = {"id": student_id}
        req_q, err = check_for_object(q_query, 'question')
        if req_q == None:
            return err

        # Remove from redis
        try:
            redis_queue = r.get("queue")
            decoded = json.loads(redis_queue)
            queue_list = decoded['queue']
            for i in range(len(queue_list)):
                if queue_list[i]['id'] == student_id:
                    queue_list.pop(i)

            decoded['queue'] = queue_list
            result = r.set("queue", json.dumps(decoded))
            if result == False:
                return handle_db_error()
        except Exception:
            return handle_db_error()

        # Send update to rabbitmq
        try:
            mq_channel.basic_publish(exchange='',
                                     routing_key=QUEUE_NAME,
                                     body="resolved")
        except (pika.exceptions.ConnectionClosed, pika.exceptions.AMQPConnectionError):
            resp = Response("RabbitMQ error", status=500, mimetype=TEXT_TYPE)
            return resp

        resp = Response("Queue updated - question resolved",
                        status=200, mimetype=TEXT_TYPE)
        return resp


# Custom error handler for status 404 - method not supported
@app.errorhandler(404)
def method_not_supported(error):
    resp = Response("404 Not Found", status=405)
    return resp

# Custom error handler for status 405 - method not supported


@app.errorhandler(405)
def method_not_supported(error):
    resp = Response("405 Method not supported", status=405)
    return resp


# Custom error handler for status 500 - internal server error
@app.errorhandler(500)
def internal_server_error(error):
    resp = Response("500 Internal Server Error", status=500)
    return resp


# check_auth checks the request for an X-User header and
# returns a 401 response if not found
def check_auth(request):
    if request.headers.get('X-User', '\{\}') == '\{\}':
        resp = Response("Unauthorized", status=401, mimetype=TEXT_TYPE)
        return resp


# Checks if the request's content type is application/json
def check_content_type(request):
    if (request.headers.get('Content-Type') != JSON_TYPE):
        resp = Response("Request body must be JSON",
                        status=415, mimetype=TEXT_TYPE)
        return resp


# Returns any errors that are caused by interaction with the database
def handle_db_error():
    resp = Response("Database error", status=500, mimetype=TEXT_TYPE)
    return resp


# Returns errors caused by an invalid JSON body
def handle_missing_field(message):
    resp = Response(message, status=415, mimetype=TEXT_TYPE)
    return resp


# Checks the database for the given type based on the query
def check_for_object(query, obj_type):
    curr_object = {}
    try:
        if obj_type == 'class':
            curr_object = db[classes].find_one(query)
        elif obj_type == 'question':
            curr_object = db[question].find_one(query)
    except pymongo.errors.PyMongoError:
        return handle_db_error()

    if curr_object == None:
        message = obj_type.capitalize() + " not found"
        resp = Response(message, status=400, mimetype=TEXT_TYPE)
        return (None, resp)

    message = obj_type.capitalize() + " already exists"
    resp = Response(message, status=400, mimetype=TEXT_TYPE)
    return (curr_object, resp)

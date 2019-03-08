from app import app
from flask import request, Response
from bson.objectid import ObjectId
from flask_pymongo import PyMongo
import pymongo
import os
import json

# MongoDB configuration
uri = os.getenv(
    "MONGO_URI", "mongodb://localhost:27017/questionqueue")
app.config["MONGO_URI"] = uri
mongo = PyMongo(app)
db = mongo.db
print(db)

classes = os.getenv("CLASS_COLLECTION", 'classes')
teachers = os.getenv("TEACHER_COLLECTION", 'teachers')
queue = os.getenv("QUEUE_COLLECTION", 'queue')

# Constants
JSON_TYPE = 'application/json'
TEXT_TYPE = 'text/plain'


# GET a current class or POST a new class
@app.route('/v1/class', methods=['GET', 'POST'])
def class_handler():
    # Check for authentication
    auth = check_auth(request)
    if auth != None:
        return auth

    if request.method == 'GET':
        # Successfully retrieves all classes; returns the encoded list in the body.
        all_classes = []
        try:
            all_classes = list(db[classes].find())
            for c in all_classes:
                c['_id'] = str(c['_id'])
        except pymongo.errors.PyMongoError:
            return handle_db_error()

        resp = Response(json.dumps(all_classes), status=200,
                        mimetype=JSON_TYPE)
        return resp
        # return 'GET /v1/class'

    elif request.method == 'POST':
        # Check content type
        content = check_content_type(request)
        if content != None:
            return content

        req_body = request.get_json()

        if req_body.get("class_number", "") == "":
            return handle_missing_field("Class number is required")
        elif req_body.get("topics", "") == "" or not isinstance(req_body['topics'], list):
            return handle_missing_field("Class topics are required")

        if not isinstance(req_body['class_number'], int):
            resp = Response("Class number must be an integer",
                            status=400, mimetype=TEXT_TYPE)
            return resp

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
        # return 'POST /v1/class'


# PATCH an existing class - overwrite topics
@app.route('/v1/class/<int:class_number>', methods=['PATCH'])
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

        # Retrieve JSON body and check for validity
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
        # return 'PATCH /v1/class'


# POST a new teacher or PATCH an authenticated user
@app.route('/v1/teacher', methods=['POST', 'PATCH'])
def teacher_handler():
    if request.method == 'POST':
        # Check for authentication
        auth = check_auth(request)
        if auth != None:
            return auth

        # Check content type
        content = check_content_type(request)
        if content != None:
            return content

        return 'POST /v1/teacher'
    elif request.method == 'PATCH':
        # Check for authentication
        auth = check_auth(request)
        if auth != None:
            return auth

        # Check content type
        content = check_content_type(request)
        if content != None:
            return content

        # Retrieve JSON body
        x_user = json.loads(request.headers.get("X-User"))
        teacher_query = {"id": x_user['id']}

        req_body = request.get_json()
        update_first = req_body.get("firstname", "")
        update_last = req_body.get("lastname", "")

        # Create update query depending on which names are to be updated
        update_query = {}
        if update_first != "" and update_last != "":
            update_query = {"$set": {"firstname": update_first,
                                     "lastname": update_last}}
        elif update_first != "":
            update_query = {"$set": {"firstname": update_first}}
        elif update_last != "":
            update_query = {"$set": {"lastname": update_last}}

        # Update and retreive updated teacher
        updated = {}
        try:
            db[teachers].update(teacher_query, update_query)
            updated = db[teachers].find_one(teacher_query)
            updated['_id'] = str(updated['_id'])
        except pymongo.errors.PyMongoError:
            return handle_db_error()

        resp = Response(json.dumps(updated), status=200, mimetype=JSON_TYPE)
        return resp
        # return 'PATCH /v1/teacher'


# GET an existing teacher
@app.route('/v1/teacher/<teacher_id>', methods=['GET'])
def specific_teacher_handler(teacher_id):
    # Check for authentication
    auth = check_auth(request)
    if auth != None:
        return auth

    if request.method == 'GET':
        teacher_query = {"id": teacher_id}

        teacher = {}
        try:
            teacher = db[teachers].find_one(teacher_query)
            teacher['id'] = str(teacher['id'])
        except pymongo.errors.PyMongoError:
            return handle_db_error()

        resp = Response(json.dumps(teacher), status=200, mimetype=JSON_TYPE)
        return resp

    # return 'GET /v1/teacher/<teacher_id>'

# DELETE a student and question from the queue
@app.route('/v1/queue/<student_id>', methods=['DELETE'])
def queue_delete_handler(student_id):
    # Check for authentication
    auth = check_auth(request)
    if auth != None:
        return auth

    if request.method == 'DELETE':
        return "queue delete handler"


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
            # print(curr_object)
    except pymongo.errors.PyMongoError:
        return handle_db_error()

    if curr_object == None:
        message = obj_type.capitalize() + " not found"
        resp = Response(message, status=400, mimetype=TEXT_TYPE)
        return (None, resp)

    message = obj_type.capitalize() + " already exists"
    resp = Response(message, status=400, mimetype=TEXT_TYPE)
    return (curr_object, resp)

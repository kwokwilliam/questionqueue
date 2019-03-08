from app import app
from flask import request


@app.route('/v1/class', methods=['GET', 'POST', 'PATCH'])
def class_handler():
    return "class handler"


@app.route('/v1/teacher', methods=['GET', 'POST', 'PATCH'])
def teacher_handler():
    return "teacher handler"


@app.route('/v1/queue/<student_id>', method=['DELETE'])
def queue_delete_handler(student_id):
    return "queue delete handler"

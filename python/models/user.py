from tortoise import fields, models
from tortoise.exceptions import ValidationError
from nicegui import ui

def validate_min(value:str):
    if(len(value) < 3):
        ui.notify('Must be 3 or more characters')
        raise ValidationError('Must be 3 or more characters')
        

class User(models.Model):
    id = fields.IntField(pk=True)
    name = fields.CharField(max_length=255,validators=[
        validate_min
    ])
    email = fields.CharField(max_length=255)


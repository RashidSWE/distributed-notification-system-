import os
from fastapi import Depends, HTTPException, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
import jwt  # uses PyJWT
from jwt import ExpiredSignatureError, InvalidTokenError

security = HTTPBearer()

JWT_SECRET = os.getenv("JWT_SECRET", "changeme")  # must match user service .env
JWT_ALGORITHM = "HS256"

def verify_jwt_token(token: str) -> dict:
    """
    Decodes JWT signed by the User Service.
    """
    try:
        payload = jwt.decode(token, JWT_SECRET, algorithms=[JWT_ALGORITHM])
        return payload
    except ExpiredSignatureError:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Token expired")
    except InvalidTokenError as e:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail=f"Invalid token: {str(e)}")


async def get_current_user(credentials: HTTPAuthorizationCredentials = Depends(security)) -> dict:
    token = credentials.credentials
    return verify_jwt_token(token)

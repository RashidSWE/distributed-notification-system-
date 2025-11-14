from sqlalchemy import create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker
import os

POSTGRES_USER = os.getenv("STATUS_DB_USER")
POSTGRES_PASSWORD = os.getenv("STATUS_DB_PASS")
POSTGRES_DB = os.getenv("STATUS_DB_NAME")
POSTGRES_HOST = os.getenv("STATUS_DB_HOST")
POSTGRES_PORT = os.getenv("STATUS_DB_PORT")

DATABASE_URL = f"postgresql://{POSTGRES_USER}:{POSTGRES_PASSWORD}@{POSTGRES_HOST}:{POSTGRES_PORT}/{POSTGRES_DB}"

engine = create_async_engine(DATABASE_URL, echo=True)
AsyncSessionLocal = sessionmaker(bind=engine, class_=AsyncSession, expire_on_commit=False)

Base = declarative_base()

class NotificationStatusEnum(str, Enum):
    delivered = "delivered"
    pending = "pending"
    failed = "failed"

class NotificationStatus(Base):
    __tablename__ = "notification_status"
    notification_id = Column(String, primary_key=True)
    status = Column(SQLAEnum(NotificationStatusEnum), nullable=False)
    error = Column(String, nullable=True)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
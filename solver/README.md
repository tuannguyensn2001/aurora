# Aurora Solver

A Python project that integrates with the z3-resolver library for constraint solving.

## Setup

### Using Virtual Environment (Recommended)

1. Create and activate virtual environment:
```bash
# Create virtual environment
python3 -m venv venv

# Activate virtual environment
source venv/bin/activate

# Or use the provided activation script
./activate.sh
```

2. Install dependencies:
```bash
pip install -r requirements.txt
```

Or install the package in development mode:
```bash
pip install -e .
```

3. Deactivate when done:
```bash
deactivate
```

### Alternative Setup (Global Installation)

```bash
pip install -r requirements.txt
```

## Usage

### Using Docker (Recommended)

```bash
# Build the image
docker build -t aurora-solver .

# Run the container
docker run -d -p 8000:8000 --name aurora-solver aurora-solver

# View logs
docker logs -f aurora-solver

# Stop and remove the container
docker stop aurora-solver
docker rm aurora-solver
```

The API will be available at http://localhost:8000

### Using Virtual Environment

Make sure your virtual environment is activated, then run:
```bash
python main.py
```

## Project Structure

```
aurora/solver/
├── venv/                 # Virtual environment (created after setup)
├── main.py              # Main application file
├── requirements.txt     # Python dependencies
├── setup.py            # Package configuration
├── activate.sh         # Convenience activation script
└── README.md           # This file
```

## API Endpoints

- `GET /` - API information and usage instructions
- `POST /solve` - Solve Z3 constraints

### Example API Usage:

```bash
# Test the API
curl http://localhost:8000/

# Solve a constraint
curl -X POST http://localhost:8000/solve \
  -H "Content-Type: application/json" \
  -d '{"constraint": "(declare-const x Int)\n(declare-const y Int)\n(assert (> x 0))\n(assert (< y 10))\n(assert (> (+ x y) 5))"}'
```

## Dependencies

- z3-solver: A constraint solver library for Python
- fastapi: Modern web framework for building APIs
- uvicorn: ASGI server for running FastAPI applications

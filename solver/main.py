from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import Optional, List
from z3 import *
import uvicorn

app = FastAPI()


class ConstraintRequest(BaseModel):
    constraint: str


class SolverResponse(BaseModel):
    check_result: str
    model: Optional[List[dict]]


@app.post("/solve", response_model=SolverResponse)
async def solve_constraint(request: ConstraintRequest):
    """
    Endpoint to solve Z3 constraints.
    
    Receives a Z3 constraint string, executes it using the solver,
    and returns the check() result and model() if satisfiable.
    """
    try:
        s = Solver()
        s.from_string(request.constraint)
        
        check_result = s.check()
        
        if check_result == sat:
            model_result = s.model()
            # Convert Z3 model to array of objects with proper types
            model_array = []
            for decl in model_result:
                val = model_result[decl]
                # Try to convert to native Python types
                if is_int(val):
                    value = val.as_long()
                elif is_bool(val):
                    value = bool(val)
                elif is_rational_value(val):
                    value = float(val.as_decimal(10))
                elif is_string_value(val):
                    value = val.as_string()
                else:
                    value = str(val)
                model_array.append({"name": str(decl), "value": value})
            return SolverResponse(
                check_result="sat",
                model=model_array
            )
        elif check_result == unsat:
            return SolverResponse(
                check_result="unsat",
                model=None
            )
        else:  # unknown
            return SolverResponse(
                check_result="unknown",
                model=None
            )
    except Exception as e:
        raise HTTPException(status_code=400, detail=f"Error processing constraint: {str(e)}")


@app.get("/")
async def root():
    return {
        "message": "Z3 Solver API",
        "usage": "POST to /solve with JSON body: {\"constraint\": \"<Z3 constraint string>\"}"
    }


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
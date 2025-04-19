-- 9. CREATE TABLE with various constraints and computed columns
CREATE TABLE employee_performance (
  evaluation_id SERIAL PRIMARY KEY,
  employee_id INTEGER NOT NULL REFERENCES employees(employee_id),
  evaluation_date DATE DEFAULT CURRENT_DATE,
  technical_score DECIMAL(3,1) CHECK (technical_score BETWEEN 0 AND 10),
  communication_score DECIMAL(3,1) CHECK (communication_score BETWEEN 0 AND 10),
  leadership_score DECIMAL(3,1) CHECK (leadership_score BETWEEN 0 AND 10),
  total_score GENERATED ALWAYS AS (technical_score + communication_score + leadership_score) STORED,
  performance_level VARCHAR(20) GENERATED ALWAYS AS (
    CASE 
      WHEN (technical_score + communication_score + leadership_score) >= 25 THEN 'Outstanding'
      WHEN (technical_score + communication_score + leadership_score) >= 20 THEN 'Exceeds Expectations'
      WHEN (technical_score + communication_score + leadership_score) >= 15 THEN 'Meets Expectations'
      ELSE 'Needs Improvement'
    END
  ) STORED,
  comments TEXT,
  CONSTRAINT unique_evaluation UNIQUE (employee_id, evaluation_date)
);


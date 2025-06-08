export interface TestCase {
  inputs: string[]
  expectedOutput: string
}

interface EngineRunBody {
  code: string
  environment: EngineEnv
  testCases: TestCase[]
}

export enum EngineRunResultStatus {
  UNKNOWN = 0,
  SUCCESS = 1,
  WRONG_ANSWER = 2,
  RUNTIME_ERROR = 3,
}

export interface EngineRunCompilationResult {
  success: boolean
  stderr?: string
}

export interface EngineRunTestCaseResult {
  inputs: string[]
  output: string
  expected_output: string
  status: EngineRunResultStatus
  error: string
  stdout: string
  execution_time: number
}

export interface EngineRunResponse {
  compilation: EngineRunCompilationResult
  results: EngineRunTestCaseResult[]
}

export async function engineRun(body: EngineRunBody) {
  const { $api } = useNuxtApp()

  return $api<EngineRunResponse>('/api/engine/run', {
    method: 'POST',
    body,
  })
}

interface EngineRunBody {
  code: string
  environment: EngineEnv
}

export interface EngineRunResult {
  execution_time: number
  exit_code: number
  stderr: string
  stdout: string
}

export async function engineRun(body: EngineRunBody) {
  const { $api } = useNuxtApp()

  return $api<EngineRunResult>('/api/engine/run', {
    method: 'POST',
    body,
  })
}

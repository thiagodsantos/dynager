import { fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ConnectionSetupPage } from '@/pages/ConnectionSetupPage'

afterEach(() => vi.unstubAllGlobals())

describe('connection setup', () => {
  it('does not ask for AWS credentials and blocks endpoint URLs containing user information', async () => {
    const fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
    render(<MemoryRouter future={{ v7_startTransition: true, v7_relativeSplatPath: true }}><ConnectionSetupPage /></MemoryRouter>)

    expect(screen.queryByLabelText(/access key|secret key|session token/i)).not.toBeInTheDocument()
    fireEvent.change(screen.getByLabelText('Connection name'), { target: { value: 'Local' } })
    fireEvent.change(screen.getByPlaceholderText('e.g. us-east-1'), { target: { value: 'us-east-1' } })
    fireEvent.change(screen.getByPlaceholderText('http://127.0.0.1:8001'), { target: { value: 'https://user:password@localhost:8001' } })
    fireEvent.click(screen.getByRole('button', { name: /test connection/i }))

    expect(await screen.findByText('Endpoint URLs cannot contain embedded usernames or passwords.')).toBeInTheDocument()
    expect(fetchMock).not.toHaveBeenCalled()
  })
})

import { describe, it, expect, vi, afterEach } from 'vitest';
import { render, screen, fireEvent } from '@solidjs/testing-library';
import { WorkItems } from './WorkItems';

const envelope = (data: unknown) => ({
  ok: true,
  status: 200,
  json: async () => ({ status: true, data, error: null }),
});

afterEach(() => vi.unstubAllGlobals());

describe('<WorkItems>', () => {
  it('renders work items fetched from the API', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        envelope([
          { id: '1', title: 'First task', project_id: 'demo', description_markdown: '', priority: 1 },
        ]),
      ),
    );

    render(() => <WorkItems />);

    expect(await screen.findByDisplayValue('First task')).toBeInTheDocument();
  });

  it('creates a work item through the API and shows it', async () => {
    const created = { id: '2', title: 'New task', project_id: 'demo', description_markdown: '' };
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(envelope([])) // initial list (empty)
      .mockResolvedValueOnce({ ok: true, status: 201, json: async () => ({ status: true, data: created, error: null }) }) // POST
      .mockResolvedValueOnce(envelope([created])); // refetch
    vi.stubGlobal('fetch', fetchMock);

    render(() => <WorkItems />);
    await screen.findByText('No work items yet. Add one above.');

    fireEvent.input(screen.getByPlaceholderText('New work item title…'), {
      target: { value: 'New task' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Add' }));

    expect(await screen.findByDisplayValue('New task')).toBeInTheDocument();

    const postCall = fetchMock.mock.calls.find(
      (c) => (c[1] as RequestInit | undefined)?.method === 'POST',
    );
    expect(postCall).toBeTruthy();
    expect(JSON.parse((postCall![1] as RequestInit).body as string)).toMatchObject({
      title: 'New task',
      project_id: 'demo',
    });
  });

  it('shows an error when the API fails', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
        json: async () => ({ status: false, data: null, error: 'boom' }),
      }),
    );

    render(() => <WorkItems />);

    expect(await screen.findByText(/boom/)).toBeInTheDocument();
  });
});

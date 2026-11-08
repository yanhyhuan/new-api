/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { afterEach, expect, test, vi } from 'vitest'

import { getBuildRevision, installBuildMetadata } from '../build-metadata'

afterEach(() => {
  vi.unstubAllEnvs()
})

test('includes the injected release version in the build revision', () => {
  vi.stubEnv('VITE_REACT_APP_VERSION', 'bao-1.17')
  expect(getBuildRevision()).toBe('rv.bao-1.17.2k6e8r7p')
})

test('uses the development fallback only when no version was injected', () => {
  vi.stubEnv('VITE_REACT_APP_VERSION', '')
  expect(getBuildRevision()).toBe('rv.0000.2k6e8r7p')
})

test('installs the same injected version in the page and runtime descriptor', () => {
  vi.stubEnv('VITE_REACT_APP_VERSION', 'bao-1.17')
  installBuildMetadata()
  expect(window.__APP_BUILD__?.rev).toBe('rv.bao-1.17.2k6e8r7p')
  expect(document.documentElement.dataset.buildRev).toBe(window.__APP_BUILD__?.rev)
  expect(document.querySelector('meta[name="build-id"]')?.getAttribute('content'))
    .toBe(window.__APP_BUILD__?.rev)
  expect(localStorage.getItem('app:rev')).toBe(window.__APP_BUILD__?.rev)
})

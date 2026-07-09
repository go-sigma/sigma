/**
 * Copyright 2023 sigma
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

export default function () {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">
      <circle cx="100" cy="100" r="100" fill="#f99d1c" />
      <circle cx="100" cy="100" r="80" fill="#cd4b27" />
      <circle cx="100" cy="100" r="60" fill="#5c2e0e" />
      <path
        d="M 50 52 L 150 52 Q 156 52 156 58 L 156 66 Q 156 72 150 72 L 104 72 Q 81 72 87 79 L 111 96 Q 120 102 111 108 L 87 125 Q 81 132 104 132 L 150 132 Q 156 132 156 138 L 156 146 Q 156 152 150 152 L 50 152 Q 44 152 44 146 L 44 140 Q 44 135 48 131 L 79 108 Q 89 102 79 96 L 48 73 Q 44 69 44 64 L 44 58 Q 44 52 50 52 Z"
        fill="#f99d1c"
        stroke="#f99d1c"
        strokeWidth="1.8"
        strokeLinejoin="round"
        strokeLinecap="round"
        transform="translate(100 100) scale(0.56) translate(-100 -102)"
      />
    </svg>
  );
}

import React from 'react';

const ExploreBg = ({ color = '#161F31', ...rest }: any) => (
  <svg width="43" height="46" xmlns="http://www.w3.org/2000/svg" {...rest}>
    <g
      stroke={color}
      fill="none"
      fillRule="evenodd"
      opacity=".4"
      strokeLinecap="round"
      strokeLinejoin="round">
      <g strokeWidth="1.04">
        <path d="M7.356 38.215v1.474-1.474zM3.811 41.76h1.474-1.474zM7.356 45.305v-1.474 1.474zM10.901 41.76H9.427h1.474z" />
      </g>
      <g
        strokeDasharray="5.695034408569335,5.695034408569335"
        strokeWidth="1.307">
        <path d="M13.526 11.778h21.222M18.628 25.888h23.375" />
      </g>
      <g strokeWidth="1.428">
        <path d="M3.301 1v.907V1zM1.12 3.181h.907-.907zM3.301 5.363v-.907.907zM5.483 3.181h-.907.907z" />
      </g>
    </g>
  </svg>
);

export default ExploreBg;

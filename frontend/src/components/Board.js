import React from 'react';

function Board({ board, onClick }) {
    return (
        <div>
            {board.map((row, i) => (
                <div key={i} style={{ display: 'flex' }}>
                    {row.map((cell, j) => (
                        <button
                            key={j}
                            onClick={() => onClick(i, j)}
                            style={{
                                width: '10vh',
                                height: '10vh',
                                fontSize: '5vh',
                                margin: '1vh'
                            }}
                        >
                            {cell}
                        </button>
                    ))}
                </div>
            ))}
        </div>
    );
}

export default Board;
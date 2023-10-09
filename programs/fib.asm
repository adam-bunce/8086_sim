bits 16

mov ax, 24 ; nth fibonacci number to find (max is 24th@46368 before overflowing)
sub ax, 1 ; might need to sub 2 depending on what you think the the "0th" number is
mov bx, 0 ; 1st num
mov cx, 1 ; 2nd num
mov dx, 0 ; hold for swapping

loop_start:
   mov dx, cx
   add cx, bx
   mov bx, dx
   sub ax, 1
   jnz loop_start
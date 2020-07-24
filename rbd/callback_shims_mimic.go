// +build !luminous

package rbd

/*

#include <rbd/librbd.h>

extern void imageWatchCallback(int index);

void callWatchCallback(int index) {
	imageWatchCallback(index);
}
*/
import "C"

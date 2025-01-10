package manager

import (
	"fmt"
	"net"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

const (
	reservedCpuCount = 2
	// reserve 2 GBs of ram
	reservedMemSize = 2147483648
)


type AvailableResource struct {
	CpuCount uint8
	CpuUtilized 	float64
	TotalMemSize uint64
	AvailableMemSize uint64
}

type Resource interface {
	GetCpuCount()	(uint8 , error)
	GetMemSize()	(uint64 , error)
	GetCpuUtilization() (float64 , error)
}

func (r AvailableResource) GetCpuCount() (uint8 , error){
	count , err := cpu.Counts(true)
	if err != nil {
		return 0 , err
	}
	if count - reservedCpuCount <= 0 {
		return 0 , nil
	}
	return uint8(count) , nil
}

func (r AvailableResource) GetMemSize() (uint64 , error) {
	v , err := mem.VirtualMemory()

	if err != nil {
		return 0 , err
	}

	if v.Available - reservedMemSize <= 0 {
		return 0 , nil
	}
	return v.Available , nil
}

func (r AvailableResource) GetCpuUtilization() (float64 , error) {
	cpuUtil , err := cpu.Percent(0  , false)
	for idx , val := range(cpuUtil) {
		fmt.Printf("%v %v \n",idx , val)
	}
	if err != nil {
		return float64(0) , err
	}
	return cpuUtil[0] , nil
}

func PrintResource(r Resource) (error){
	cpu , err := r.GetCpuCount()
	if err != nil {
		return err
	}
	ram , err := r.GetMemSize()
	if err != nil {
		return err
	}
	cpuUtil , err := r.GetCpuUtilization()
	if err != nil {
		return err
	}
	fmt.Printf("CPU COUNT: %v \n" , cpu)
	fmt.Printf("CPU UTIL: %v \n" , cpuUtil)
	fmt.Printf("RAM: %v \n" , ram)
	return nil
}

func CheckPortAvailable(port string) (error) {
	listener , err := net.Listen("tcp",":" + port)
	if err != nil {
		fmt.Printf("%v is busy\n",port)
		return err
	}
	listener.Close()
	fmt.Printf("%v is open",port)
	return nil
}